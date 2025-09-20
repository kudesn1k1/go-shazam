package core

import (
	"go-shazam/internal/core/db"

	"go.uber.org/fx"
)

var Module = fx.Module("core",
	fx.Provide(db.NewDB, db.NewTransactionManager, db.NewRepository),
)
