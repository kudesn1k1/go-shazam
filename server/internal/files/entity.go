package files

import "time"

const (
	StatusTemporary = "TEMPORARY"
	StatusConfirmed = "CONFIRMED"
)

type FileEntity struct {
	Hash        string    `db:"hash"`
	ContentType string    `db:"content_type"`
	SizeBytes   int       `db:"size_bytes"`
	Status      string    `db:"status"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

func ObjectKey(hash string) string {
	return "files/" + hash
}
