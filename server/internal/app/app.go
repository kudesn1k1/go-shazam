package app

import (
	"context"
	"go-shazam/internal/core"
	"go-shazam/internal/fingerprint"
	appHttp "go-shazam/internal/http"
	"go-shazam/internal/queue"
	"go-shazam/internal/recognition"
	"go-shazam/internal/song"
	"go-shazam/internal/spotify"
	"go-shazam/internal/youtube"
	"net/http"

	"go.uber.org/fx"
)

func NewWebApp() *fx.App {
	return fx.New(
		core.Module,
		fingerprint.Module,
		appHttp.Module,
		// Core modules
		song.Module,
		spotify.Module,
		youtube.Module,
		recognition.Module,
		queue.Module,
		// Http modules
		song.HttpModule,
		recognition.HttpModule,

		fx.Invoke(core.RegisterCoreMiddleware),
		fx.Invoke(func(r *http.Server) {}),
	)
}

func NewWorkerApp() *fx.App {
	return fx.New(
		core.Module,
		fingerprint.Module,
		song.Module,
		spotify.Module,
		youtube.Module,
		recognition.Module,
		queue.Module,
		fx.Invoke(registerWorkerLifecycle),
	)
}

func registerWorkerLifecycle(lc fx.Lifecycle, w queue.WorkerServer) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			return w.Start()
		},
		OnStop: func(ctx context.Context) error {
			w.Stop()
			return nil
		},
	})
}
