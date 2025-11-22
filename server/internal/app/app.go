package app

import (
	"go-shazam/internal/core"
	"go-shazam/internal/fingerprint"
	appHttp "go-shazam/internal/http"
	"go-shazam/internal/recognition"
	"go-shazam/internal/song"
	"go-shazam/internal/spotify"
	"go-shazam/internal/youtube"
	"net/http"

	"go.uber.org/fx"
)

func NewApp() *fx.App {
	return fx.New(
		core.Module,
		fingerprint.Module,
		appHttp.Module,
		song.Module,
		spotify.Module,
		youtube.Module,
		recognition.Module,
		fx.Invoke(func(r *http.Server) {}),
	)
}
