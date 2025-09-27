package core

import (
	"go-shazam/internal/core/db"
	"go-shazam/internal/core/middleware"

	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
)

var Module = fx.Module("core",
	fx.Provide(db.NewDB, db.NewTransactionManager, db.NewRepository),
	fx.Invoke(func(r *gin.Engine) {
		r.Use(middleware.CorrelationMiddleware())
	}),
)
