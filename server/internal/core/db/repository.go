package db

import (
	"context"
)

type Repository struct {
	tm *TransactionManager
}

func NewRepository(tm *TransactionManager) *Repository {
	return &Repository{tm: tm}
}

func (r *Repository) Connection(ctx context.Context) Executor {
	return r.tm.GetConnection(ctx)
}
