package app

import (
	"go-shazam/internal/core"
	"go-shazam/internal/http"

	"go.uber.org/fx"
)

func NewApp() *fx.App {
	return fx.New(
		core.Module,
		http.Module,
	)
}
