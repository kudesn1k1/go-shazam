package app

import (
	"go-shazam/internal/core"
	appHttp "go-shazam/internal/http"
	"go-shazam/internal/song"
	"go-shazam/internal/spotify"
	"go-shazam/internal/youtube"
	"net/http"

	"go.uber.org/fx"
)

func NewApp() *fx.App {
	return fx.New(
		core.Module,
		appHttp.Module,
		song.Module,
		spotify.Module,
		youtube.Module,
		fx.Invoke(func(r *http.Server) {}),
	)
}
