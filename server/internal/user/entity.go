package user

import (
	"time"

	"github.com/google/uuid"
)

type UserEntity struct {
	ID             uuid.UUID `db:"id"`
	Email          string    `db:"email"`           // Encrypted email
	EmailHash      string    `db:"email_hash"`      // SHA-256 hash for lookups
	HashedPassword string    `db:"hashed_password"` // bcrypt hash
	CreatedAt      time.Time `db:"created_at"`
	UpdatedAt      time.Time `db:"updated_at"`
}
