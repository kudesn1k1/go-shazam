package app

import (
	"context"
	"fmt"
	"go-shazam/internal/auth"
	"go-shazam/internal/core"
	"go-shazam/internal/files"
	"go-shazam/internal/fingerprint"
	appHttp "go-shazam/internal/http"
	"go-shazam/internal/queue"
	"go-shazam/internal/recognition"
	"go-shazam/internal/role"
	"go-shazam/internal/song"
	"go-shazam/internal/spotify"
	"go-shazam/internal/user"
	"go-shazam/internal/web"
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
		role.Module,
		song.Module,
		spotify.Module,
		youtube.Module,
		recognition.Module,
		queue.Module,
		user.Module,
		files.Module,
		web.Module,
		// Http modules
		song.HttpModule,
		recognition.HttpModule,
		user.HttpModule,
		files.HttpModule,
		web.HttpModule,

		fx.Invoke(core.RegisterCoreMiddleware),
		fx.Invoke(registerTestEndpoints),
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
		files.Module,
		song.QueueModule,
		files.CleanupModule,
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
