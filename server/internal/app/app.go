package app

import (
	"context"
	"fmt"
	"go-shazam/internal/auth"
	"go-shazam/internal/core"
	"go-shazam/internal/fingerprint"
	appHttp "go-shazam/internal/http"
	"go-shazam/internal/queue"
	"go-shazam/internal/recognition"
	"go-shazam/internal/song"
	"go-shazam/internal/spotify"
	"go-shazam/internal/user"
	"go-shazam/internal/youtube"
	"net/http"

	"go.uber.org/fx"
)

func NewWebApp() *fx.App {
	return fx.New(
		core.Module,
		fingerprint.Module,
		appHttp.Module,
		auth.Module,
		// Core modules
		song.Module,
		spotify.Module,
		youtube.Module,
		recognition.Module,
		queue.Module,
		user.Module,
		// Http modules
		song.HttpModule,
		recognition.HttpModule,
		user.HttpModule,

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
		song.QueueModule,
		fx.Invoke(registerWorkerLifecycle),
	)
}

func registerWorkerLifecycle(lc fx.Lifecycle, w queue.WorkerServer) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			if err := w.Start(); err != nil {
				fmt.Printf("Failed to start worker: %v\n", err)
				return err
			}
			fmt.Println("Worker server started successfully")
			return nil
		},
		OnStop: func(ctx context.Context) error {
			fmt.Println("Stopping worker server...")
			w.Stop()
			return nil
		},
	})
}
