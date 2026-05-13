// Package fixtures loads SQL fixtures and supports between-test cleanup.
package fixtures

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
)

// Load applies a fixture .sql file (by basename, e.g. "songs-30.sql") against db.
// The file must live in server/test/fixtures/.
func Load(t *testing.T, db *sqlx.DB, name string) {
	t.Helper()
	_, thisFile, _, _ := runtime.Caller(0)
	path := filepath.Join(filepath.Dir(thisFile), name)

	sqlBytes, err := os.ReadFile(path)
	require.NoError(t, err)

	_, err = db.Exec(string(sqlBytes))
	require.NoError(t, err)
}

// Truncate wipes the named tables in dependency order. Use between subtests
// inside one suite; the testcontainers helpers handle inter-suite cleanup.
func Truncate(t *testing.T, db *sqlx.DB, tables ...string) {
	t.Helper()
	for _, table := range tables {
		_, err := db.Exec("TRUNCATE TABLE " + table + " RESTART IDENTITY CASCADE")
		require.NoError(t, err)
	}
}
