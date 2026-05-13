// Package containers boots ephemeral Docker containers for integration tests.
// Use Start* helpers from inside `//go:build integration` test files only.
package containers

import (
	"context"
	"database/sql"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/pressly/goose/v3"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

// StartPostgres boots an ephemeral Postgres container, runs every goose
// up-migration from server/migrations against it, and returns an open
// *sqlx.DB. Cleanup runs automatically via t.Cleanup.
func StartPostgres(t *testing.T) *sqlx.DB {
	t.Helper()
	ctx := context.Background()

	c, err := tcpostgres.Run(ctx,
		"postgres:17-alpine",
		tcpostgres.WithDatabase("test"),
		tcpostgres.WithUsername("test"),
		tcpostgres.WithPassword("test"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(60*time.Second),
		),
	)
	require.NoError(t, err)
	t.Cleanup(func() { _ = c.Terminate(context.Background()) })

	dsn, err := c.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	require.NoError(t, runMigrations(dsn))

	db := sqlx.MustOpen("pgx", dsn)
	t.Cleanup(func() { _ = db.Close() })
	return db
}

func runMigrations(dsn string) error {
	rawDB, err := sql.Open("pgx", dsn)
	if err != nil {
		return err
	}
	defer rawDB.Close()

	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}
	return goose.Up(rawDB, findMigrationsDir())
}

// findMigrationsDir resolves the server/migrations path regardless of CWD.
func findMigrationsDir() string {
	_, thisFile, _, _ := runtime.Caller(0)
	// thisFile = .../server/test/containers/postgres.go
	return filepath.Clean(filepath.Join(filepath.Dir(thisFile), "..", "..", "migrations"))
}
