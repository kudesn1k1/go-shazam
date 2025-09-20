package db

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"go.uber.org/fx"
)

type txKey struct{}

type TransactionManager struct {
	db *sqlx.DB
}

type Executor interface {
	Exec(query string, args ...any) (sql.Result, error)
	Get(dest interface{}, query string, args ...interface{}) error
	Select(dest interface{}, query string, args ...interface{}) error
}

func NewTransactionManager(db *sqlx.DB) *TransactionManager {
	return &TransactionManager{db: db}
}

func NewDB(lc fx.Lifecycle) *sqlx.DB {
	config := LoadDBConfig()
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", config.Host, config.Port, config.User, config.Password, config.Database)
	fmt.Println(dsn)
	db := sqlx.MustOpen("pgx", dsn)

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			fmt.Println("DB connected")
			return db.PingContext(ctx)
		},
		OnStop: func(ctx context.Context) error {
			fmt.Println("DB disconnected")
			return db.Close()
		},
	})

	return db
}

func Transactional[T any](ctx context.Context, tm *TransactionManager, fn func(ctx context.Context) (T, error)) (T, error) {
	var zero T

	tx, err := tm.db.BeginTx(ctx, nil)
	if err != nil {
		return zero, err
	}

	ctxWithTx := context.WithValue(ctx, txKey{}, tx)

	res, err := fn(ctxWithTx)
	if err != nil {
		_ = tx.Rollback()
		return zero, err
	}

	if err := tx.Commit(); err != nil {
		return zero, err
	}

	return res, nil
}

func (tm *TransactionManager) GetConnection(ctx context.Context) Executor {
	if tx, ok := ctx.Value(txKey{}).(*sqlx.Tx); ok {
		return tx
	}
	return tm.db
}
